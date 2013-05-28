import java.awt.color.CMMException;
import java.awt.image.BufferedImage;
import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.ArrayList;
import java.util.HashMap;

import javax.imageio.ImageIO;

/**
 * This program detects and deletes duplicate images in a folder by comparing MD5 hashes
 * arg[i] must be absolute path to folder and encased with quotes if it contains spaces 
 * i.e "\home\Zeroxk\My Pictures"
 * 
 * @author Zeroxk
 *
 */

public class ImgDupDeleter {

	static ArrayList<String> imgExts = new ArrayList<String>();
	static int totalDupes;

	@SuppressWarnings("unchecked")
	public static void main(String[] args) {

		if(args.length < 1) {
			System.out.println("No folders specified");
			System.exit(0);
		}

		imgExts.add(new String("png"));
		imgExts.add(new String("jpg"));
		imgExts.add(new String("jpeg"));
		imgExts.add(new String("gif"));

		for (int i = 0; i < args.length; i++) {

			File folder = new File(args[i]);

			if(folder.isDirectory()) {
				try {
					
					StringBuilder sb = new StringBuilder();
					sb.append(folder.getAbsolutePath());
					sb.append(File.separator);
					sb.append(folder.getName());
					sb.append(".txt");
					File hashes = new File(sb.toString());

					HashMap<String, File> mapFiles = new HashMap<>();
					
					if(hashes.createNewFile() || hashes.length() == 0) {
						System.out.println("Created file for storing hashes " + hashes.getName());
					}else {
						System.out.println(hashes.getName() + " already exists, loading\n");
						FileInputStream fs = new FileInputStream(hashes);
						ObjectInputStream in = new ObjectInputStream(fs);

						mapFiles = (HashMap<String, File>) in.readObject();
						System.out.println("Loaded existing hashmap " + hashes.getName());
						in.close();
						fs.close();
					}
					
					checkFolder(folder, mapFiles);
					
					//Serialize mapFiles and store
					FileOutputStream fs = new FileOutputStream(hashes);
					ObjectOutputStream out = new ObjectOutputStream(fs);
					out.writeObject(mapFiles);
					out.close();
					fs.close();
					System.out.println("Serialized and stored hashmap " + hashes.getName());
					
				} catch (FileNotFoundException e) {
					e.printStackTrace();
				} catch (ClassNotFoundException e) {
					e.printStackTrace();
				} catch (IOException e) {
					e.printStackTrace();
				}
			}else {
				System.out.println(args[i] + " is not a folder");
			}

		}
		
		System.out.println("Total number of dupes: " + totalDupes);

	}

	/**
	 * Checks a folder for duplicates, does recursive call if the file being worked on is a folder
	 * @param folder Folder to be checked
	 * @throws FileNotFoundException 
	 * @throws IOException 
	 * @throws ClassNotFoundException 
	 */
	private static void checkFolder(File folder, HashMap<String, File> mapFiles) throws FileNotFoundException, ClassNotFoundException, IOException {

		System.out.println("Processing folder: " + folder.getAbsolutePath());
		File [] images = folder.listFiles();
		int numDupes = 0;

		for (File currFile : images) {
			
			if(currFile.isDirectory()) {
				checkFolder(currFile, mapFiles);
				continue;
			}
			
			if(mapFiles.containsValue(currFile)) {
				System.out.println(currFile.getAbsolutePath() + " has already been checked, skipping\n");
				continue;
			}			

			String name = currFile.getName();
			String imgExt = name.substring(name.lastIndexOf(".")+1);

			if(!imgExts.contains(imgExt)) {
				System.out.println(name + " ignored, not an image\n");
				continue;
			}

			System.out.println("Processing image: " + currFile.getAbsolutePath());

			byte[] hash = hash(currFile, imgExt);
			if(hash == null) {
				System.out.println(currFile.getAbsolutePath() + " could not be hashed, deleted\n");
				continue;
			}
			String hex = hashToHex(hash);

			if(mapFiles.containsKey(hex) && !mapFiles.get(hex).equals(currFile)) {
				System.out.println(currFile.getAbsolutePath() + " is duplicate of " + mapFiles.get(hex).getAbsolutePath());
				numDupes++;
				totalDupes++;
				currFile.delete();
			}else {
				mapFiles.put(hex, currFile);
			}
		}

		System.out.println("Number of dupes in folder " + folder.getAbsolutePath() + ": " + numDupes);
		System.out.println("Done with folder: " + folder.getAbsolutePath());

	}


	/**
	 * Transforms MD5 hash to hex string
	 * @param hash Digest to be converted
	 * @return Hex string of the hash
	 */
	private static String hashToHex(byte[] hash) {
		String str = "";

		for (int i = 0; i < hash.length; i++) {
			str += Integer.toString( (hash[i] & 0xff) + 0x100, 16).substring(1);
		}

		return str;
	}

	/**
	 * Hashes an image using MD5
	 * @param img Image to be hashed
	 * @return Returns byte array of the hash
	 */
	private static byte[] hash(File img, String imgExt) {

		byte[] hash = null;
		try {


			BufferedImage buffImg = ImageIO.read(img);
			if(buffImg == null) return null;
			ByteArrayOutputStream outputStream = new ByteArrayOutputStream();
			ImageIO.write(buffImg, imgExt, outputStream);
			byte[] data = outputStream.toByteArray();

			System.out.println("Starting MD5 digest");
			MessageDigest md = MessageDigest.getInstance("MD5");
			md.update(data);
			hash = md.digest();

		} catch (IOException e) {
			System.out.println("Error while processing file as image");
			e.printStackTrace();
		} catch (NoSuchAlgorithmException e) {
			System.out.println("Error: could not find MD5 algorithm");
			e.printStackTrace();
		} catch (CMMException e) {
			System.out.println("Could not load image");
			e.printStackTrace();
		} catch (ArrayIndexOutOfBoundsException e) {
			System.out.println("Weirdo bug when reading certain gifs");
			e.printStackTrace();
		}

		System.out.println("Done with digest of image: " + img.getAbsolutePath() + "\n");
		return hash;

	}

}
